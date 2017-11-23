# Dracarys
Bring fire and fury to your all too peaceful containers

## TL;DR

Done with unit testing and integration testing of your container-based app?

Time to bring fire and fury to your containers and see what remains of your app!

Dracarys integrates a set of cool Go libraries (Docker, Pumba, Vegeta) to bring chaos to your system, by:

- messing with your containers *e.g.* restarting random container
- messing with the network *e.g.* delay or lose packet
- load your endpoints *e.g.* POSTing high number of HTTP requests
- administrating your system *e.g.* scaling service up/down, rolling out updates, changing resources

## So, what new?

Well, you might think you'd be better up directly using Docker, Pumba and Vegeta. They are indeed great tools.

Dracarys allows you to specify chaos scenarios that are:

- easy to specify,
- but not too naive
- repeatable,
- yet with a degree of random
- easy to correlate with measured metrics

## Used Libraries

- Docker API
- Pumba. As the original Pumba is an app, I am currently using my fork, which can be used as a lib.
- Vegeta (soon) or maybe Gatling as an alternative/complement. Wait, no, Gatling is implemented in Scala...

## License

Released under Apache License v2
