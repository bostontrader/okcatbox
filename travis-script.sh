# 1. Using the demo Bookwerx server, get credentials and setup some demo data.
BSERVER="http://185.183.96.73:3003"
APIKEY="$(curl -X POST $BSERVER/apikeys | jq -r .apikey)"

CURRENCY_BTC="$(curl -d "apikey=$APIKEY&rarity=0&symbol=BTC&title=Bitcoin" $BSERVER/currencies | jq .LastInsertId)"
CURRENCY_LTC="$(curl -d "apikey=$APIKEY&rarity=0&symbol=LTC&title=Litecoin" $BSERVER/currencies | jq .LastInsertId)"

HOT_WALLET_BTC="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_BTC&title=Hot wallet" $BSERVER/accounts | jq .LastInsertId)"
HOT_WALLET_LTC="$(curl -d "apikey=$APIKEY&rarity=0&currency_id=$CURRENCY_LTC&title=Hot wallet" $BSERVER/accounts | jq .LastInsertId)"

CAT_ASSETS="$(curl -d "apikey=$APIKEY&symbol=A&title=Assets" $BSERVER/categories | jq .LastInsertId)"
CAT_LIABILITIES="$(curl -d "apikey=$APIKEY&symbol=L&title=Liabilities" $BSERVER/categories | jq .LastInsertId)"
CAT_EQUITY="$(curl -d "apikey=$APIKEY&symbol=Eq&title=Equity" $BSERVER/categories | jq .LastInsertId)"
CAT_REVENUE="$(curl -d "apikey=$APIKEY&symbol=R&title=Revenue" $BSERVER/categories | jq .LastInsertId)"
CAT_EXPENSES="$(curl -d "apikey=$APIKEY&symbol=Ex&title=Expenses" $BSERVER/categories | jq .LastInsertId)"
CAT_FUNDING="$(curl -d "apikey=$APIKEY&symbol=F&title=Funding" $BSERVER/categories | jq .LastInsertId)"
CAT_HOT="$(curl -d "apikey=$APIKEY&symbol=H&title=Hot wallet" $BSERVER/categories | jq .LastInsertId)"

curl -d "apikey=$APIKEY&account_id=$HOT_WALLET_BTC&category_id=$CAT_ASSETS" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$HOT_WALLET_LTC&category_id=$CAT_ASSETS" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$HOT_WALLET_BTC&category_id=$CAT_HOT" $BSERVER/acctcats
curl -d "apikey=$APIKEY&account_id=$HOT_WALLET_LTC&category_id=$CAT_HOT" $BSERVER/acctcats

# 2. Build a config file...
echo "bookwerx:" > okcatbox.yaml
echo "  apikey: $APIKEY" >> okcatbox.yaml
echo "  server: $BSERVER" >> okcatbox.yaml
echo "  funding: $CAT_FUNDING" >> okcatbox.yaml
echo "  hot: $CAT_HOT" >> okcatbox.yaml
echo "listenaddr: :8090" >> okcatbox.yaml

# 3. Crank it up!
./okcatbox -config=okcatbox.yaml &

# 4. Now execute okprobe commands against the okcatbox server
CATBOX_URL=http://localhost:8090

# 4.1 Get Catbox credentials.  Recall that these credentials are unique to the Catbox.
CATBOX_CREDENTIALS=okcatbox-read.json
curl -X POST $CATBOX_URL/catbox/credentials --output $CATBOX_CREDENTIALS
CB_APIKEY="$(cat okcatbox-read.json | jq -r .api_key)"

# 4.2 Simulate the deposit of coin into the funding account.
curl -d "apikey=$CB_APIKEY&currency_symbol=BTC&quan=0.5&time=2020-07-21T" $CATBOX_URL/catbox/deposit

# funding
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt currencies
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt deposit-address
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt deposit-history
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt wallet
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt withdrawal-fee

# spot
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt accounts