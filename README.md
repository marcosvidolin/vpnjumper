# VPN Jumper

A simple HTTP foward

## Motivation

To access certain network addresses (services and APIs), it was necessary to use a VPN. The problem was that the VPN client does not work on my Mac.
It's easier to do this than to install a VM to run Windows with the VPN.

## How to Run

Server

```
go run . -type=server
```

Client

```
go run . -type=client
```

## Integration

The integration is done through Redis.

https://app.redislabs.com/
