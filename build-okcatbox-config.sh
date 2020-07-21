# For some reason this doesn't work directly from .travis.yml
echo "bookwerx:" > okcatbox.yaml
echo "  apikey: $APIKEY" >> okcatbox.yaml
echo "  server: $BSERVER" >> okcatbox.yaml
echo "  funding: $CAT_FUNDING" >> okcatbox.yaml
echo "  hot: $CAT_HOT" >> okcatbox.yaml
echo "listenaddr: :8090" >> okcatbox.yaml