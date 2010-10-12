#!/bin/sh

set -e

if [ ! -n "$1" ]
then
  echo
  echo "Usage: `basename $0` benchmark-name [additional benchmark args]"
  echo
  echo "Benchmarks:"
  for x in `find . -name "*.go" | xargs basename | cut -d "." -f 1`; do
     echo "-" $x
  done
  echo
  echo "You must have Go installed (http://golang.org)"
  exit 1
fi

cd `dirname $0`

APP_NAME=$1

TEMP_DIR=`mktemp -d`

GO_FILE=$1.go

6g -o $TEMP_DIR/out.6 $GO_FILE 

6l -o $TEMP_DIR/$APP_NAME $TEMP_DIR/out.6

shift

$TEMP_DIR/$APP_NAME $@

rm -r $TEMP_DIR