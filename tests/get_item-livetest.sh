curl -H "X-Bbpd-Indent: true" -H "X-Amz-Target: DynamoDB_20120810.GetItem" -X POST -d '{"TableName":"test-godynamo-livetest","Key":{"TheHashKey":{"S":"a-hash-key-json1"},"TheRangeKey":{"N":"1"}}}' "http://localhost:12333/?compact=1&indent=1";
echo "";
curl -H "X-Bbpd-Verbose: True" -H "X=BBPD-Verbose: True" -X POST -d '{"TableName":"test-godynamo-livetest","Key":{"TheHashKey":{"S":"a-hash-key-json1"},"TheRangeKey":{"N":"1"}}}' "http://localhost:12333/GetItemJSON?compact=1";

