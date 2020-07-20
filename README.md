[![Build Status](https://travis-ci.org/bostontrader/okcatbox.svg?branch=master)](https://travis-ci.org/bostontrader/okcatbox)
[![MIT license](http://img.shields.io/badge/license-MIT-brightgreen.svg)](http://opensource.org/licenses/MIT)

# Welcome to the OKCatbox

OKEx provides an API for using their service.  Unfortunately, learning how to use the real OKEx API looks suspiciously close to DOS and general hackery from their point of view.  And heaven forbid you actually launch the missiles by accident while you learn to use their API.  

With the OKCatbox we attempt to simulate the real OKEx server.  This is of course easier said than done, but we've tried.  We will discuss divergences from reality shortly.

OKEx provides a testnet of its own, but according to its docs, said testnet is limited to "futures and options contract".  At this time, the OKCatbox focuses on spot trading, so it fills a need that OKEx doesn't already provide.

## Getting Started

### The Easy Way

The easiest way to get started is to use our online demo located at http://185.183.96.73:8090.  Using any tool of your choice, create the HTTP requests that you would ordinarily send to the real OKEx API but instead send them to the demo OKCatbox server.  One particularly helpful tool for this task is [OKProbe](http://github.com/bostontrader/okprobe). In fact, we will assume the use of OKProbe in these instructions, so please install this first.

As with the real OKEx server, you'll need credentials for most of the API calls.  You can get them from the OKCatbox server:

```
export CATBOX_URL=http://185.183.96.73:8090
export CATBOX_CREDENTIALS=okcatbox.json
curl -X POST $CATBOX_URL/catbox/credentials --output $CATBOX_CREDENTIALS
```

WARNING! [Danger!](https://www.youtube.com/watch?v=1IPPn9t6dyE).  When using the OKCatbox, especially when using our demo server, DO NOT use your real OKEx credentials!

### The Hard Way

You can always install the OKCatbox server yourself.

As prerequisites, you'll need git and golang installed on your system as well as access to a [Bookwerx Core](https://github.com/bostontrader/bookwerx-core-rust) server.  Assuming you have these things:

```
git clone https://github.com/bostontrader/okcatbox
cd okcatbox
go build
./okcatbox -help
```

Next, you'll need to configure the OKCatbox to use bookwerx-core.

The first step in doing that is to find a suitable bookwer-core server.  We have a [publicly available demonstration version](http://185.183.96.73:3003) for your convenience.  This is the URL for submitting RESTful commands.  We also provide a [a convenient UI](http://185.183.96.73:3005/) that works with this server.

We're going to use the URL of the bookwerx-core server in subsequent requests, so let's save it in the env:
```
export BSERVER=http://185.183.96.73:3003
```

Next, we need to obtain an API key for the bookwerx-core server.  We can do this using the aforementioned UI, but the command-line way is:

```
curl -X POST $BSERVER/apikeys
```

The response looks intuitively obvious to our human eyeballs.  However, we're going to need to use this value repeatedly in the future, so it would be really useful to parse this response, pick out just the value of the apikey, and save just that value in the env.  We can easily parse this using [jq](https://stedolan.github.io/jq/). Assuming jq is properly installed and combining all this new-found learning into one command yields:

```
export APIKEY="$(curl -X POST $BSERVER/apikeys | jq -r .apikey)"
```

Next, we have to define some currencies.  Any currency that the OKCatbox can deal with will have to be defined in the bookwerx-core server. Let's start this party by defining Bitcoin and Litecoin.  As usual, you can do this with the UI or use the command line:

```
export CURRENCY_BTC="$(curl -d "apikey=$APIKEY&rarity=0&symbol=BTC&title=Bitcoin" $BSERVER/currencies | jq .LastInsertId)"
export CURRENCY_LTC="$(curl -d "apikey=$APIKEY&rarity=0&symbol=LTC&title=Litecoin" $BSERVER/currencies | jq .LastInsertId)"
```

Upon close inspection you can see a parameter named "rarity".  It's harmless but not relevant for this tutorial so just ignore it.

Another wrinkle is that we have double quotes inside double quotes.  Oddly enough this seems to work for us, but this looks like something that might go wrong for somebody else, so be wary of this.


Now we have to define a handful of accounts.  We will need a hot wallet for each of the above currencies.

```
export HOT_WALLET_BTC="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=Hot wallet" $BSERVER/accounts | jq .LastInsertId)"
export HOT_WALLET_LTC="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_LTC&title=Hot wallet" $BSERVER/accounts | jq .LastInsertId)"
```
Even though the titles of these accounts are the same, they are distinguished via the different currencies.

As with currencies, we have the rarity parameter that is just as irrelevant here as there and equally ignorable.

Next, we'll need to define some useful categories that we can use to tag the accounts:

|Symbol |Title       |
|-------|------------|
|A	    |Assets      | 
|L      |Liabilities |
|Eq	    |Equity      |
|R      |Revenue     |
|Ex	    |Expenses    |
|F	    |Funding     |
|H      |Hot wallet  |

There are many cases where we need to find accounts that are somehow related.  For example, a Balance Sheet will need to identify all Assets.

The Funding category is used to tag any user account related to newly deposited funds in the funding account.

The Hot wallet category is used to tag an accounts that are hot wallets.

Even though we don't presently have accounts that are members of all these categories, we'll need them later.  Let's just get this over with now.

```
export CAT_ASSETS="$(curl -d "apikey=$APIKEY&symbol=A&title=Assets" $BSERVER/categories | jq .LastInsertId)"
export CAT_LIABILITIES="$(curl -d "apikey=$APIKEY&symbol=L&title=Liabilities" $BSERVER/categories | jq .LastInsertId)"
export CAT_EQUITY="$(curl -d "apikey=$APIKEY&symbol=Eq&title=Equity" $BSERVER/categories | jq .LastInsertId)"
export CAT_REVENUE="$(curl -d "apikey=$APIKEY&symbol=R&title=Revenue" $BSERVER/categories | jq .LastInsertId)"
export CAT_EXPENSES="$(curl -d "apikey=$APIKEY&symbol=Ex&title=Expenses" $BSERVER/categories | jq .LastInsertId)"
export CAT_FUNDING="$(curl -d "apikey=$APIKEY&symbol=F&title=Funding" $BSERVER/categories | jq .LastInsertId)"
export CAT_HOT="$(curl -d "apikey=$APIKEY&symbol=H&title=Hot wallet" $BSERVER/categories | jq .LastInsertId)"
``` 

Now let's tag the accounts with suitable categories. 

| Category | Account       | Currency|
|----------|---------------|---------|
| A	       | Hot Wallet    | BTC     |
| A	       | Hot Wallet    | LTC     |
| H	       | Hot Wallet    | BTC     |
| H	       | Hot Wallet    | LTC     |

This is pretty simple right now.  The two hot wallets, one for each of the two currencies, are both assets.

```
curl -d "apikey=$APIKEY&account_id=$HOT_WALLET_BTC&category_id=$CAT_ASSETS" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$HOT_WALLET_LTC&category_id=$CAT_ASSETS" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$HOT_WALLET_BTC&category_id=$CAT_HOT" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$HOT_WALLET_LTC&category_id=$CAT_HOT" $BSERVER/acctcats
```
In these cases, even though we still make http requests to do this, we don't care about saving any information from the responses.

Finally, let's create a configuration file for the OKCatbox.

Note: Be aware that the PWD was irrelevant for the prior commands, but it affects the location of the okcatbox.yaml file created next. 

````
echo "bookwerx:" > okcatbox.yaml
echo "  apikey: $APIKEY" >> okcatbox.yaml
echo "  server: $BSERVER" >> okcatbox.yaml
echo "  funding: $CAT_FUNDING" >> okcatbox.yaml
echo "  hot: $CAT_HOT" >> okcatbox.yaml
echo "listenaddr: :8090" >> okcatbox.yaml

````

Note that the listenaddr can be set to any usable port.


And now... [Drumroll please...](https://www.youtube.com/watch?v=-R81ugVBLlw&t=9)

./okcatbox -config=okcatbox.yaml

The OKCatbox server is now running and listening on whatever port you specified earlier.


## Divergences from Reality

There are a handful of divergences from reality regarding how closely the OKCatbox can mimic the real OKEx server.  Said divergences include:

* The OKCatbox has an extra API call to provide the credentials needed by most of the real OKEx API calls. Said credentials are only usable by the OKCatbox and _are not_ usable on the real OKEx server.  Good luck stealing coins with them.

* The OKCatbox has an extra API call to enable the user to assert a deposit into his funding account.  Ordinarily, deposits work by sending coins to the real OKEx server and waiting for it to notice.  This is impractical for the OKCatbox.  So we just make the assertion and be done with it.

* The OKCatbox can be accessed via http.  (Note: The real OKEx server can also be accessed via http even though their docs specify https.)

* The http response headers returned from the real server, as well as their contents, are obscure and mostly undocumented. Our testing looks for their presence, makes some guesses about values, as well as looking for extra headers.  If your app depends upon using any of this, you'll need to take a closer look.

* Similarly, the http request headers sent by the client require a handful of documented examples, but we have not tried to spam the server with extraneous headers.
