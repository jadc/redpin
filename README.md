# 📌 redpin
A Discord app for pinning messages and comparing stats, powered by Golang.

This repository contains the modern rewrite in Go. If you are looking for the original python version, see [redpin-py](https://github.com/jadc/redpin-py). I rewrote this bot to gather some practice dealing with APIs, routing, and asynchronous events in the Go programming language.

Unlike most pin bots, redpin is highly configurable and flexible, supporting any number of emojis to pin messages; in addition, the bot uses webhooks for a much more minimal appearance of saved messages.

## Tour
TODO

## Contribution

Install [Nix](https://nixos.org/download) and then run `nix develop` in the root of the repo.

Redpin does not have many dependencies, so if you do not want to install Nix, you can read the [flake.nix](flake.nix) file and figure out what you need.
