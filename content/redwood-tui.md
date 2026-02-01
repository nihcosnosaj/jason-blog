---
title: "What's that above me?"
date: 2026-01-30
slug: "redwood-tui-above-me"
---

This weekend's project was something I've been wanting to tackle for a while, mostly inspired by nerdy obsession with aviation. I wanted to know what was flying up above and around me. Sure, I could check out FlightAware or FlightRadar24 in a browser, but I feel they have become bloated, monetized, and visually noisy. Inspired by the [hardware-centric builds](https://blog.colinwaddell.com/articles/flight-tracker) of others, I tried to build something that I could boot up on a mini screen or in a corner of my monitor and it would run in the background (with less hardware overhead, as cool as Colin's build is). 

And so **Redwood** was born: a high-performance, minimalist TUI (text/terminal user interface) for real-time aircraft telemetry closest to you. 

I looked at `bubbletea` in Go, but as I've been loving doing work in Rust recently, I chose the ruthless memory safety and concurrency primitives of Rust's `ratatui` ecosystem. 

And yes, the name `redwood` is a nod to the callsign for Virgin America. Sorely missed from the skies of the U.S., but not forgotten. Devoured by Alaska Airlines. 

### Architecture
In `ratatui`, unlike Elm-inspired libraries like Go's `bubbletea`, state management is ... unopionionated. The orchestration of threads and managing state is completely up to you. Aiming for 60 FPS while making API calls async in the background and querying a local SQLite db, this is the architecture I ironed out:
1. a main loop -- a central loop that handles drawing the UI and dealing with user input. 
2. the async poller -- a background `tokio` task managing the OpenSky API lifecyle.
3. the blocking worker -- a dedicated pool for CPU-bound database enrichment.

### Scaling Metadata
OpenSky provides a nifty (and free) endpoint at `/states/all` that provides raw telemetry and barebones data based on a geographical (latitute and longitutde) area you define in the request. The response is quite light, but gives a meaningful starting point with the following metadata:
- `icao24` -- a unique ICAO 24-bit address of the transponder in hex string representation
- `callsign`, `origin_country`, `timestamp`, `velocity`, `vertical_rate`, `true_track` (heading in degrees)

While this is cool, it doesn't really help me to populate a dashbord or title card for showing planes near me. I want to know what airline, what aircraft type, etc.

After digging, I found OpenSky also has a 500MB CSV of metadata on every `icao24` with much, much more info. Bingo. I grabbed fields related to owner, operator, operator callsign, aircraft manufacturer and model so that I could attach `United Airlines, Boeing 787-900` to the live aircraft I got back from the API response. 

I simply load this CSV into a local SQLite database when the app boots up for the first time. This achieved two very beneficial things:
1. Lookup times for airline and aircraft type were very quick, usually less than 1 millisecond.
2. We didn't have to worry about another external API call, for now. 

Granted, we still have no info on something I really want - origin airport and destination airport. That is one of the main things still not implemented, but I fear that feature lives behind a monthly subscription that I'm not ready to muster the funds to get. 

### Systems-Level Concurrency: MPSC and Arc
Passing state into `tokio::spawn` was a tricky thing to learn and fully grasp. I initially toyed with wrapping the `App` state in an `Arc<Mutex<App>>`, but the "lock contention" in a TUI can be quite deadly if not handled with care. If the background thread holds the lock while the UI tries to draw, we get dropped frames. So I opted to share memory by communicating, like a true Rustacean :) 

I used MPSC (Multi-Producer, Single-Consumer) channels. Very cool stuff. My `EventHandler` owns the reciever, while the background API and DB poller clones the sender:
```Rust

// src/events.rs
pub struct EventHandler {
    pub tx: mpsc::UnboundedSender<Event>,
    rx: mpsc::UnboundedReceiver<Event>,
}

impl EventHandler {
    pub fn new(tick_rate_ms: u64) -> Self {
        let (tx, rx) = mpsc::unbounded_channel(); 
        // excluded code here...
    }
}

// src/main.rs
// Background API Poller
let api_tx = events.tx.clone(); // Clone the handle, not the state
tokio::spawn(async move {
    loop {
        if let Ok(flights) = provider.fetch_overhead(user_lat, user_lon, radius).await {
            // Offload enrichment to prevent blocking the async executor
            let enriched = tokio::task::spawn_blocking(move || db::decorate_flights(flights))
                .await
                .unwrap_or_default();

            let _ = api_tx.send(Event::FlightUpdate { flights: enriched, ... });
        }
        tokio::time::sleep(Duration::from_secs(poll_interval)).await;
    }
});
```
We clone the sender, and move that to a background API task. Then, the poller can send `FlightUpdate` events back to the UI and the UI doesn't even have to ask for them. This sort of message passing ensures we have zero stuttering on the main UI rendering and that the UI thread is the only source of truth for app state. Background tasks simply "suggest" updates to the main thread (UI) by sending these async messages. No deadlocks!

### Oh the Math
One thing I had to figure out was how to sort flights by closest to our geolocated coordinates. Since we are searching in a circle area, I found the [Haversine formula](https://en.wikipedia.org/wiki/Haversine_formula) to take care of that. Casting the trigonometry into `f64` types keep it as accurate as possible. Then we just sort the `Vec<Flight>` before each render cycle via Rust's [`partial_cmp`](https://doc.rust-lang.org/std/cmp/trait.PartialOrd.html) method. 

### Next Steps

There is a lot I want to do on this still. This is more of a to-do list for myself that I came up with during development, but here are some other things I want to tack on to this:
- DB connection pooling
    - I'm opening a new SQLite connection each batch of flights. This is quite expensive. 
- Zero-Copy JSON parsing 
    - `serde` is great, don't get me wrong, but I'm currenlty deserializing into an intermediate `Value` map instead of directly into a struct. 
    - `simd-json` seems *interesting* for this, but needs some more playing with on my part. I'd hope it would reduce allocation pressure on the CPU and give us lower latency in the poller loop.
- Optimize the `profile.release` section to reduce the binary size. Smaller the footprint, the better!
- Error handling (and in that same vein, logging statements to the log file) need to be more robust and detailed.
- Resilience around hitting the API endpoint. What if it returns a 429 (Rate Limit) or a 503 (Overloaded), but the loop keeps trying to hit it? Maybe if a request fails to get an OK response, we double the `poll_interval_seconds` temporarily? We don't want OpenSky to blacklist our IP, and always want to be respectful of such awesome, free, open-source resources.
- Typing feels messy to me. Does each field require its own type? (e.g. lat, lon, altitude, velocity, etc)
- More graceful shutdown protocol!
- Offline mode? Caching of last position incase the app is opened without internet?
- Some sort of filtering mechanism. What if I only care about planes passing at crusing altitude? or if I finally buy into Alaska's product and become a diehard Alaska flyer?
- Benchmarking in Github actions to avoid performance regressions with new changes or features. 
- Just caving and turning it into a GUI? The lack of a map is bothering me, and ASCII maps just don't look good. 
    - In that vein, a radar view would be cool. The center being your location, complete with a "sweeping" bar around the circle that populates new aircraft or updates their position. Color coded for descending/ascending/stable?
- Altitude map of the flights?
- Ghost trails? Storing the last 5-10 positions for each aircraft in a HashMap and rendering them out with fading intenstity. 
- I'd love to get the airline logo somehow, but I think color might be all we can do in a shell. 





