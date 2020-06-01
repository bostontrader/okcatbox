# 1. Build a config file...

# 1.1 ... using the demo Bookwerx server.
BSERVER="http://185.183.96.73:3003"
APIKEY="$(curl -X POST $BSERVER/apikeys | jq -r .apikey)"
echo "bookwerx:" >> okcatbox.yaml
echo "  apikey: $APIKEY" >> okcatbox.yaml
echo "  server: $BSERVER" >> okcatbox.yaml

# 1.2 ... set a listening address:port for the OKCatbox
echo "listenaddr: :8090" >> okcatbox.yaml

# 2. Execute the OKCatbox using the prior created config
./okcatbox -config=okcatbox.yaml &

# 3. Execute okprobe commands against the okcatbox server
CATBOX_URL=http://localhost:8090

# These credentials are unique to the Catbox
CATBOX_CREDENTIALS=okcatbox-read.json

# funding
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt currencies
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt deposit-address
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt deposit-history
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt wallet
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt withdrawal-fee

# spot
okprobe -url $CATBOX_URL -errors -keyfile $CATBOX_CREDENTIALS -endpnt accounts