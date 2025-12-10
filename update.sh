git pull
cd backend
docker build -t ms_backend .
cd ..
cd frontend
docker build -t ms_frontend .
cd ..
docker compose down
docker compose up -d