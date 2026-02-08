---
title: "Towards Chaos"
date: 2026-02-07
slug: "towards-chaos-thomas-attractor"
---

<img src="/assets/thomas-attractor.gif" alt="chaos-demo.gif" style="width:100% !important; height: auto;">

All things eventually trend toward chaos and decay. Which is funny because we humans spend our lives trying to build and preserve institutions of order. Spreadsheets, calendars, daily to-do lists, enrolling in a school program, etc. 

Chaos doesn't mean order isn't present, though. It means there was a very sensitive dependence on initial conditions. There is a particularly interesting branch of mathematics, Chaos Theory, that I spent some time reading about and wanted to explore visually. [Read.](https://en.wikipedia.org/wiki/Chaos_theory) Some more reading later and I stumbled upon the Thomas Cyclically Symmetrical Attractor, a quite poetic map of how a dynamic system can descend into chaos. 

Most systems want to settle to rest. Pendulums stop, a cup of coffee levels off to room temperature, a ball kicked by a person will roll for a bit but eventually come back to rest. Chaotic systems, on the other hand, are inherently attracted to a state of perpetual, non-repeating motion. 

The Thomas Attractor traces a path through 3D space governed by a set of three cyclical differential equations, that is they all cycle back to influence one another:

$$
\begin{aligned}
\frac{dx}{dt} &= \sin(y) - bx \\\\
\frac{dy}{dt} &= \sin(z) - by \\\\
\frac{dz}{dt} &= \sin(x) - bz
\end{aligned}
$$

It wonderfully creates a labyrinth of soft curves that is visually unique and beautiful. 

The dissipation constant, $b$ , acts as a drag, trying to pull the system to a standstill. The sine functions are the constant injection of energy and oscillation. When we lower the constant, the lines start to become interweaving and lattice-like. They'll never cross themselves and never end. 

I think chaos is often mistaken as a lack of control, but the Thomas Attractor and the math behind say differently. It's quite wonderful that a system can be governed by rigid laws and yet remain full of surprise and novelty. It's always tracing a path that has never been walked before. 

### Implmentation in Rust

As expected, I modeled it in Rust. Opted for a Euler integration for the visual look of it. I couldn't get the RK4 to look how I wanted it to. A great example of added complexity not always benefitting the system. 

I used the `eframe` and `egui` crates so I could eventually try porting it to WASM with near-native performance. That part is still a work-in-progress, but it renders beautifully in a desktop GUI. 

