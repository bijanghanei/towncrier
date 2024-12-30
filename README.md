# TOWNCRIER

This project consists of two microservices , repoter and publisher. The reporter miroservice has the resposiblity to fetch news from X (formely twitter) and send them in kafka partions related to each username. The publisher microservice has the responsiblity to filter the tweets in the kafka topic and partions based on specific words and publish them to any telegram channel that has @towncrier telegram bot.

## Setup

1. Replace `"YOUR_BEARER_TOKEN"` in `docker-compose.yml` with your actual X API bearer token.
2. Ensure Docker and Docker Compose are installed on your machine.
3. Build and run the application with Docker Compose:

```sh
docker-compose up --build