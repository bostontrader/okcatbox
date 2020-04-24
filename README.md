# Welcome to the OKCatbox

OKEx provides an API for using their service.  Unfortunately, learning how to use the real API looks suspiciously close to DOS and general hackery from their point of view.  And heaven forbid you actually launch the missiles by accident while you learn to use the API.  

With the OKCatbox we attempt to simulate the real OKEx server.  This is of course easier said than done, but we've tried.  We will discuss divergences from reality shortly.

OKEx provides a testnet of its own, but according to its docs, said testnet is limited to "futures and options contract".  At this time, the Catbox focuses on spot trading so it fills a need that OKEx doesn't already provide.

## Divergences from Reality

There are a handful of divergences from reality regarding how closely the OKCatbox can mimic the real OKEx server.  Said divergences include:

* The OKCatbox has an extra API call to provide the credentials needed by most of the real OKEx API calls. Said credentials are only usable by the OKCatbox and _are not_ usable on the real OKEx server.

* The OKCatbox has an extra API call to brainwipe the server.

* The OKCatbox can be accessed via http.  (Note: The real OKEx server can also be accessed via http even though their docs specify https.)

* The http response headers returned from the real server, as well as their contents, are obscure and mostly undocumented. Our testing looks for their presence, makes some guesses about values, as well as looking for extra headers.  If your app depends upon using any of this, you'll need to take a closer look.

* Similarly, the http request headers sent by the client require a handful of documented examples, but we have not tried to spam the server with extraneous headers.
