# RATE LIMIT FOR REQUEST HTTPS

## HOW TO RUN
1. ./gen_certs.sh
2. Copy key in certs folder into file .env
3. make run

## HOW TO TEST
Call request to address https://localhost:4000, server just accept 10 requests within 10 seconds. If exceed server will response message with "Rate limit error" text.
