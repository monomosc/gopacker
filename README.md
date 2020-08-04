# gopacker
A way to get Vendored Modules into Corporate Networks

Currenty Runs under https://gopacker.monomo.network

Example call: `https://gopacker.monomo.network/package?package=https://github.com/micromdm/scep`

This will take awhile and then spit out a `package.zip` containing the Repo as if you used these:

```
git clone <Package>
go mod vendor
zip -r package.zip .
```

This is useful for me because at work I'm sitting behind several layers of Proxies and Firewalls - they won't let `git` or `go` talk. But my browser is allowed to do what he wants...


Notes: I have no idea what happens if you call my API with a Repo that does not exist or which is not a Go Repo. Probably HTTP 500.

Last note: Please don't kill my service, I'll just have to add BasicAuth.
