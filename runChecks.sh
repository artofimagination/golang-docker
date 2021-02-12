go clean -testcache && go test -v -count=1 -cover ./...
status=$?; 
if [[ $status != 0 ]]; then 
  exit $status; 
fi
golangci-lint run -v .
status=$?; 
if [[ $status != 0 ]]; then 
  exit $status; 
fi
./runFunctionalTest.sh
status=$?; 
if [[ $status != 0 ]]; then 
  exit $status; 
fi