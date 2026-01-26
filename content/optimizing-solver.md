---
title: "Optimizing my Grid-Based Solver"
date: 2026-01-25
slug: "optimizing-the-solver"
---

![fluidsim_demo](/assets/fluidsim_demo.gif)

A few months ago, I spent a weekend building out a solver for the Navier-Stokes equations for an incompressible fluid in Rust. I had to brush up on the math needed which I haven't touched since university, but the main goal of the project was to be pushed off the deep end on Rust and emerge a Rustacean. Did I succeed on that front? Maybe. I still have a long way to go on fully grasping Rust (and it's strange yet powerful concepts like ownership) like I do Python or Go, but I feel much more confident in the language.

Anyway, I felt the need to revisit this project and try to make some optimizations. The actual simulation is quite clunky, and the codebase felt messy to me as I lumped everything into main.

Some goals for this were:
- Improve the performance by some measurable metric. 
- Along with that, eliminate any major bottlenecks within the solver code.
- Refactor the codebase to seperate the Model from the View as is a staple of good, clean code. 
- Apply some more advanced Rust concepts I've read about since initial implementation. (Mainly things i've learned about while reading Klabnik's *The Rust Programming Language*)

### Performance Optimization

The main way I wanted to improve the performance was through how I was allocating memory.

On my initial implementation, I was foolishly allocating 50 new vectors each iteration via `.clone()` inside loops like so:
```Rust

    // Step 2: Pressure Solve
    // Solves the Poisson equation to enforce incompressibility.
    fn solve_pressure(&mut self, dt: f64) {
        // Implementation of the Jacobi iteration for the pressure solve.
        // This is where we calculate divergence and iterate to find p.
        let dx = self.dx;
        // We can assume density is 1 for simplicity, as it scales the pressure
        let rho = 1.0; 

        // Part 1: Calculate the divergence of the velocity field.
        // This is the right-hand side of the Poisson equation.
        let mut divergence = vec![0.0; self.nx * self.ny];
        for j in 0..self.ny {
            for i in 0..self.nx {
                let u_right = self.u[self.u_idx(i + 1, j)];
                let u_left  = self.u[self.u_idx(i, j)];
                let v_top   = self.v[self.v_idx(i, j + 1)];
                let v_bot   = self.v[self.v_idx(i, j)];

                let d = (u_right - u_left + v_top - v_bot) / dx;
                divergence[self.p_idx(i, j)] = d;
            }
        }

        // Part 2: Iteratively solve for pressure using the Jacobi method.
        // We repeat this loop to let the pressure values settle.
        let mut p_new = self.p.clone();
        let num_iterations = 50; // More iterations = more accuracy
        for _ in 0..num_iterations {
            for j in 1..self.ny - 1 { // We only solve for interior pressure points
                for i in 1..self.nx - 1 {
                    let p_right = self.p[self.p_idx(i + 1, j)];
                    let p_left  = self.p[self.p_idx(i - 1, j)];
                    let p_top   = self.p[self.p_idx(i, j + 1)];
                    let p_bot   = self.p[self.p_idx(i, j - 1)];

                    let d = divergence[self.p_idx(i, j)];

                    // This is the discretized Poisson equation rearranged for p_i,j
                    let p_updated = (p_right + p_left + p_top + p_bot - d * dx * dx) / 4.0;
                    p_new[self.p_idx(i, j)] = p_updated;
                }
            }
            // Update the pressure field for the next iteration
            self.p = p_new.clone();


```

As one might imagine, this creates quite a performance bottleneck as we allocate a new `Vec` each loop. I did some digging around the web and found the idea of double buffering --> we keep two persistent buffers and swap in between them. This forum [post](https://users.rust-lang.org/t/thread-safe-double-buffer-implementation/81693) was helpful as was some asking around Gemini. 

To actually implement this, I introduced a few "previous" state buffers (`p_prev`, `u_prev`, and `v_prev`), as well as a scratch buffer (`divergence`) to my `FluidGrid` struct. So now, instead of calling `.clone()` each loop, these "previous" bufferes are reused without allocating any extra memory. 

To handle the swapping between them, I used `std::mem::swap` combined with `copy_from_slice` to copy data without reallocation. 

Here is how it looks with the changes:

```Rust

    // Step 2: Pressure Solve
    // Solves the Poisson equation to enforce incompressibility.
    fn solve_pressure(&mut self, _dt: f64) {
        // Implementation of the Jacobi iteration for the pressure solve.
        // This is where we calculate divergence and iterate to find p.
        let dx = self.dx;

        // Part 1: Calculate the divergence of the velocity field.
        // This is the right-hand side (RHS) of our Poisson equation.
        for j in 0..self.ny {
            for i in 0..self.nx {
                let u_right = self.u[self.u_idx(i + 1, j)];
                let u_left  = self.u[self.u_idx(i, j)];
                let v_top   = self.v[self.v_idx(i, j + 1)];
                let v_bot   = self.v[self.v_idx(i, j)];

                let d = (u_right - u_left + v_top - v_bot) / dx;
                let idx = self.p_idx(i, j);
                self.divergence[idx] = d;
            }
        }

        // Part 2: Iteratively solve for pressure using the Jacobi method.
        // We repeat this loop to let the pressure values settle.
        // Copy current p to p_prev to start with a good guess
        self.p_prev.copy_from_slice(&self.p);
        let num_iterations = 50; // More iterations = more accuracy
        for _ in 0..num_iterations {
            for j in 1..self.ny - 1 { // We only solve for interior pressure points
                for i in 1..self.nx - 1 {
                    let p_right = self.p[self.p_idx(i + 1, j)];
                    let p_left  = self.p[self.p_idx(i - 1, j)];
                    let p_top   = self.p[self.p_idx(i, j + 1)];
                    let p_bot   = self.p[self.p_idx(i, j - 1)];

                    let d = self.divergence[self.p_idx(i, j)];

                    // This is the discretized Poisson equation rearranged for p_i,j
                    let p_updated = (p_right + p_left + p_top + p_bot - d * dx * dx) / 4.0;
                    let idx = self.p_idx(i, j);
                    self.p_prev[idx] = p_updated;
                }
            }
            // Update the pressure field for the next iteration
            mem::swap(&mut self.p, &mut self.p_prev);
        }     

    }
```

As stated, I wanted a way to measure this improvement on performance. To do this meaningfully, I put in a "slow" mode that I could toggle during runtime. I kept the `.clone()` logic alongside and measured (in microseconds) the time to compute each frame of the sim and print to stdout. I created a simple toggle switch (`B` on the keyboard during simulation). After doing that, I ran the simulation and switched back and forth. Below, you can see quite a decrease in compute time per frame:

```Shell
Optimized mode: true
Frame time: 104.042µs
Frame time: 106.041µs
Frame time: 98.625µs
Frame time: 105.917µs
Frame time: 98.5µs
Frame time: 103.125µs
Frame time: 110.209µs
Frame time: 99.833µs
Frame time: 104.083µs
Optimized mode: false
Frame time: 175.208µs
Frame time: 168.875µs
Frame time: 168.708µs
Frame time: 182.708µs
Frame time: 187.125µs
Frame time: 171.875µs
Frame time: 171.417µs
Frame time: 172.25µs
```

Based on this, we can extract a bit of concrete metrics on how much we increased our performance through double buffering. We got a roughly 69% increase in performance (or a roughly 41% reduction in compute time per frame). Pretty cool stuff!

- Optimized Average: ~103.38 µs per frame
- Unoptimized Average: ~174.77 µs per frame


### Decoupling the Visualization from the Simulation

The secondary objective of all this was to clean up the code bit and seperate concerns. I moved the rendering/view code to its own structure. 

The `FluidRenderer` struct purely handles coloring the speed gradient, and drawing the two modes I currenlty have implemented: vector field and scalar heatmap. 

This now keeps the `FluidGrid` struct to handling just the physics simulation / solving aspect of things. 

### Final Organization Tweaks

As I was moving the `FluidRenderer` logic into its rightful struct, I realized I should probably just conform to the MVC pattern and move everything to its own file, so I did that and added the following:
- `grid.rs` --> to house `FluidGrid`
- `renderer.rs` --> to house `FluidRenderer`


### Closing

There is still quite a lot I want to do with this solver to get it more to a water/smoke sim state, both in performance and visually. I would love to explore using `rayon` to parallelize it in a better way to decrease the compute time per frame even more. 

We are currently using semi-lagrangian advection with bilinear interpolation, which is stable, but the trade off is that we are smoothing out a lot of the interesting swirls and details that make water and smoke sims so interesting to generate. Some topics I've read a bit about but need to find a way to implement include: vorticity confinement, advecting a density field, and maybe swap out the semi-lagrangian advection for MacCormack advection to get a sharper edge on the fluid. 

Also, now that we are optimized, I'd like to see how it handles higher resolutions like 100x100.

Maybe next weekend!


