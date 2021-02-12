docker-compose down
docker stop $(docker ps -aq)
docker rm $(docker ps -aq)
docker system prune -f
docker-compose up --build --force-recreate -d main-server
python3 -m pytest -v test