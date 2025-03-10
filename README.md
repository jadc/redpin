# 📌 redpin

A [Discord app](https://support.discord.com/hc/en-us/articles/21334461140375-Using-Apps-on-Discord#h_01J7CJ994TKKMGYMG1ZZQ9T3S5) for saving messages and comparing stats, powered by the [Go programming language](https://go.dev).

Unlike most pin bots, redpin is highly configurable and flexible, supporting any number of emojis to pin messages; in addition, the bot uses webhooks and modern Discord features for much more minimal aesthetics compared to other options.

## Tour

TODO

## Motive

This repository contains the modern rewrite and expansion of **redpin** in the Go. If you are looking for the original Python version, see [redpin-py](https://github.com/jadc/redpin-py); I rewrote this bot to practice interfacing with databases and APIs, routing packets, responding to asynchronous events, and generally extend the features of this bot significantly past its current form and existing (free and paid) offerings.

## Installation
The recommended way of hosting is using [Docker](https://www.docker.com/).
1. Clone this repository.
   
   ```sh
   git clone https://github.com/jadc/redpin.git
   ```
2. Create an environment file named `.env` with the following contents.
   
   ```
   DISCORD_TOKEN=<discord token here>
   DB_FILE=[optional, path to database file]
   ```
3. Run the container using Docker Compose.
   
   ```sh
   docker compose up
   ```

Of course, you can host it without Docker. If you want to do it that way, you probably know how. :)

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
    - [x] Pins preserve attachments when possible
    - [x] Leaderboard for users with most pins (and emojis used)
- [ ] Write tests
- [x] Write CI/CD pipeline for automatic building and running tests
- [x] Write Dockerfile for production deployment
- [x] Write Nix Shell for development environment
- [ ] Showcase bot in [Tour](#Tour)
