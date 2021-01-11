docker-compose down
docker system prune -f
docker-compose up --build --force-recreate -d main-server
python3 -m pytest -v test