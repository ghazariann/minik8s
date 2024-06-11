#!/bin/bash

# Create the DNS resource using kubectl and capture the output
output_create=$(./kubectl create dns -f testdata/dns.yaml)
expected_output_create="Dns created: test-dns"

# Check if the DNS was created successfully
if [[ "$output_create" == "$expected_output_create" ]]; then
    echo "DNS creation output matched expected output."
else
    echo "DNS creation failed or output did not match."
    echo "Expected: $expected_output_create"
    echo "Got: $output_create"
    exit 1
fi

# Test the DNS by curling the domain
output_curl=$(curl -s vahag.com)
expected_content="<title>My Website</title>"

# Check if the curl response is as expected
if [[ "$output_curl" == *"$expected_content"* ]]; then
    echo "DNS is correctly resolving and the web page content is as expected."
else
    echo "Failed to retrieve correct web page content."
    echo "Expected to contain: $expected_content"
    echo "Got: $output_curl"
    exit 1
fi

# Clean up the DNS resource after the test
cleanup_output=$(./kubectl delete dns test-dns)
expected_cleanup_output="Dns deleted: test-dns"

if [[ "$cleanup_output" == "$expected_cleanup_output" ]]; then
    echo "DNS cleanup was successful."
else
    echo "Failed to clean up the DNS resource."
    echo "Expected: $expected_cleanup_output"
    echo "Got: $cleanup_output"
    exit 1
fi

echo "All DNS checks passed!"
