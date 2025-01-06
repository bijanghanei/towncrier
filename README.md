# TOWNCRIER

This project consists of two microservices , repoter and publisher. The reporter miroservice has the resposiblity to fetch news from X (formely twitter) and send them in kafka partions related to each username. The publisher microservice has the responsiblity to filter the tweets in the kafka topic and partions based on specific words and publish them to any telegram chat that has @towncrierrbot telegram bot.

## Setup

1. Replace `"TELEGRAM_BOT_TOKEN"` in `docker-compose.yml` with your actual TELEGRAM BOT token.
2. Replace `"X_BEARER_TOKEN"` in `docker-compose.yml` with your actual X API token.
3. Replace the username that you want to listen to (`"X_USERNAMES"` in `docker-compose.yml`).
4. Ensure Docker and Docker Compose are installed on your machine.
5. Build and run the application with Docker Compose:

```sh
docker-compose up -d
