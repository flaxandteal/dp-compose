#!/usr/local/bin/bash
url=http://localhost:8082/login
curl -d '{"email":"florence@magicroundabout.ons.gov.uk","password":"<your password here>"}' $url