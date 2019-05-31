#!/bin/bash
# In response to long-time issue: https://github.com/golang/go/issues/2381
# This is a barely modified version of this gist by github user "Kegsay"
# Found here: https://gist.github.com/Kegsay/84ce060f237cb9ab4e0d2d321a91d920
set -u

DOC_DIR=godoc
PKG=github.com/smartedge/codechallenge

# Run a godoc server which we will scrape.
godoc -http=localhost:6060 &
DOC_PID=$!

# Wait for the server to init
echo -n "Waiting for godoc server"
while :
do
    curl -f -s "http://localhost:6060/pkg/$PKG/" > /dev/null
    if [ $? -eq 0 ] # exit code is 0 if we connected
    then
        break
    fi
    echo -n .
done
echo

# Scrape the pkg directory for the API docs. Scrap lib for the CSS/JS. Ignore everything else.
# The output is dumped to the directory "localhost:6060".
wget -r -m -k -E -np -p -erobots=off --include-directories="/pkg,/lib" \
	--exclude-directories="*" "http://localhost:6060/pkg/$PKG/"

# Stop the godoc server
kill -9 $DOC_PID

# Delete the old directory or else mv will put the localhost dir into
# the DOC_DIR if it already exists.
rm -rf $DOC_DIR
mv localhost\:6060 $DOC_DIR

echo "Docs can be found in $DOC_DIR"
echo "Replace /lib and /pkg in the gh-pages branch to update gh-pages"
