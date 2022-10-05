#!/bin/bash
set -x
awslocal s3 mb s3://dp-interactives-file-uploads
awslocal s3 mb s3://deprecated
awslocal s3 mb s3://testing
awslocal s3 mb s3://testing-public
set +x