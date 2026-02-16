---
title: "Homelab Hero"
date: 2026-02-15
slug: "homelab-hero"
---

My homelab journey has been ... rough. I've been trying to procure some NUCs or tiny-mini-micro PCs for a bit now for a fair price via marketplace or ebay, but haven't had any luck. 

I first pivoted to building a Go wrapper on top of kubernetes-in-docker (KinD) to spin environments up and down locally. It worked, but it felt like eating at Olive Garden when I wanted Italy. I wanted real networking, real instance lifecycles, and real stakes.

So I spent some time playing with Terraform and refreshing my AWS skills to build out a pretty nifty little virtual homelab with a strict budget. 

A lean, mean, kubernetes machine. :) 

### Why didn't you just use EKS?
EKS is an enterprise powerhouse, but for one engineer in an SF apartment built in 1907, it’s overkill. EKS charges a flat hourly fee for the control plane ($0.10/hr) before you even add a single node. That doesn't quite align with my mission for this, plus it abstracts away a lot of the fun of learning the nitty gritty. 

### Motivation
My motivation for this was threefold:
1. **Dive deep into the primitives of AWS.** I wanted to dive into the weeds on EC2, VPC, and handling all the networking stuff at a granular level. It's easy to abstract away too much and not know what to do when things break in production.
2. **Keep costs to a minimum.** My budget is tiny. I'm aiming for under $30 per month for active use on this.
3. **Go.** Go is probably my favorite language to write code in. I wanted another excuse to use it, so I used it to help me manage my nodes and deployments by interacting with AWS directly, managing the lifecycle of my virtual infra. 

### Costs
Behold, the Spot Instance. These are a lifesaver for a budget homelab hero (title mention). They are unused cloud capacity that are offered at very deep discounts as compared to regular compute. By using AWS’s spare capacity, I can get a `t4g.xlarge` (4 vCPUs, 16GB RAM) for roughly $0.066/hr. About $48/month if ran 24/7. 

As one can expect, they are not persistent and are prone to 2-minute warnings that AWS wants that compute back. 

If I need state, I'm eyeballing Longhorn or S3/RDS for internal AWS solutions. Would love to also play around with Rook or Ceph, but definitley overkill for these tiny ARM nodes. 

### Go Manager
I wanted some way to manage these and get some CLI output on what I was doing cost-wise. I built `nebula-manager` in Go that fetches the actual live Spot Instance price for our AWS resource type in our region (`us-west-1a`) directly from the AWS `DescribeSpotPriceHistory` API so that when I boot down my homelab via CLI, it outputs the estimated cost. 

There is a lot to add on to this tooling for more functionality, but it does work quite well as a bare bones manager for tearing down and cost reciepts. 

### Next Steps
I'm going to run it for the next month and see what the costs amount to, while expanding the Go tooling to be more of a robust watchdog. 

Also want to deploy the Prometheus helm chart on the cluster and try to set up some alerts based on spending and resource availability. 



