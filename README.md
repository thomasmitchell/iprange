# iprange

A program that can check if a given IP in within an IP range. Does stuff that
you can already do with `sipcalc`, but with an emphasis on simplicity and easier
scriptability.

```text
 range <target> <minrange> <maxrange>
    Check if IP is between two other ips

  cidr <target> <range>
    Check if an ip is in a CIDR range

  convert <range>
    Takes a CIDR and gives the min and max IP address in the range, newline separated

  subtract <minuend> <subtrahend>
    Subtract one CIDR from another and get the resulting range(s)
```

## Obtaining things and stuff

With go installed, you can use `go get -u github.com/thomasmmitchell/iprange` to
get the newest code and a fresh build. Alternatively, you can go to the Github
releases page and pick up a binary just for you.

Use `iprange --help` for more info on usage.