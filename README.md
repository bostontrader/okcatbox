[![Build Status](https://travis-ci.org/bostontrader/okcatbox.svg?branch=master)](https://travis-ci.org/bostontrader/okcatbox)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

# Welcome to the OKCatbox

OKEx provides an API for using their service.  Unfortunately, learning how to use the real API looks suspiciously close to DOS and general hackery from their point of view.  And heaven forbid you actually launch the missiles by accident while you learn to use their API.  

With the OKCatbox we attempt to simulate the real OKEx server.  This is of course easier said than done, but we've tried.  We will discuss divergences from reality shortly.

OKEx provides a testnet of its own, but according to its docs, said testnet is limited to "futures and options contract".  At this time, the OKCatbox focuses on spot trading so it fills a need that OKEx doesn't already provide.

## Getting Started

The easiest way to get started is to use our online demo located at http://185.183.96.73:8090.  Using any tool of your choice, create appropriate HTTP requests and send them to said server.  One particularly helpful tool for this task is [OKProbe](http://github.com/bostontrader/okprobe). In fact, we will assume the use of OKProbe in these instructions.

1. Install [OKProbe](http://github.com/bostontrader/okprobe)

2. The OKCatbox has a beginner set of credentials hardwired in.  Create a text file in a location of your choice.  Assuming it's named okcatbox.json:

{ "api_key" : "47477ba4-74ad-4649-4c71-36c587a82c7d"
, "api_secret_key" : "4790CA744289696413598ECBAB430B79"
, "passphrase" : "valid passphrase"
}

Please be advised that these credentials are shared by anybody who sees them and that they're only valid for an OKCatbox.  

3. Submit requests to the OKCatbox by using OKProbe.  For example:

./okprobe -url http://185.183.96.73:8090 -endpnt wallet -keyfile  /path/to/okcatbox.json

4. You can get your own set of credentials at http://185.183.96.73:8090/credentials

Please also be advised that the OKCatbox uses an in-memory db that is brainwiped upon restart.  If your credentials don't work that's probably the cause.  Just get new credentials.

## Divergences from Reality

There are a handful of divergences from reality regarding how closely the OKCatbox can mimic the real OKEx server.  Said divergences include:

* The OKCatbox has an extra API call to provide the credentials needed by most of the real OKEx API calls. Said credentials are only usable by the OKCatbox and _are not_ usable on the real OKEx server.  Good luck stealing coins with them.

* The OKCatbox can be accessed via http.  (Note: The real OKEx server can also be accessed via http even though their docs specify https.)

* The http response headers returned from the real server, as well as their contents, are obscure and mostly undocumented. Our testing looks for their presence, makes some guesses about values, as well as looking for extra headers.  If your app depends upon using any of this, you'll need to take a closer look.

* Similarly, the http request headers sent by the client require a handful of documented examples, but we have not tried to spam the server with extraneous headers.
