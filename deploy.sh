cd aahs-func 

sam package --output-template-file packaged.yaml --s3-bucket aahs-golang

sam deploy --template-file packaged.yaml --stack-name aahs-golang --capabilities CAPABILITY_IAM