# 📌 redpin

A [Discord app](https://support.discord.com/hc/en-us/articles/21334461140375-Using-Apps-on-Discord#h_01J7CJ994TKKMGYMG1ZZQ9T3S5) for saving messages and comparing stats, powered by the [Go programming language](https://go.dev).

Unlike most pin bots, redpin is highly configurable and flexible, supporting any number of emojis to pin messages; in addition, the bot uses webhooks and modern Discord features for much more minimal aesthetics compared to other options.

## Tour

TODO

## Motive

This repository contains the modern rewrite and expansion of **redpin** in the Go. If you are looking for the original Python version, see [redpin-py](https://github.com/jadc/redpin-py); I rewrote this bot to practice interfacing with databases and APIs, routing packets, responding to asynchronous events, and generally extend the features of this bot significantly past its current form and existing (free and paid) offerings.

## TODO

- [x] Implement core functionality
    - [x] Message detection
    - [x] Automatic webhook creation and usage
    - [x] Message pinning
    - [x] Message pinning with replies
    - [x] Command-based interface for per-guild configuration
        - [x] Pin Channel
        - [x] Reaction Count to Pin (each or sum)
        - [x] Allow pins from NSFW channels
        - [x] Allow self-pins
        - [x] Select which emojis pin messages
- [ ] Implement nice-to-haves
    - [ ] Emoji-themed pin channels
    - [ ] Per-user leaderboard for how many of their messages have been pinned (toggleable)
    - [ ] Pin score: ratio between pins given and pins received (toggleable)
- [ ] Write tests
- [x] Write CI/CD pipeline for automatic building and running tests
- [ ] Write Dockerfile for production deployment
- [x] Write Nix Shell for development environment
- [ ] Showcase bot in [Tour](#Tour)
