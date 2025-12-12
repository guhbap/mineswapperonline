#!/bin/bash

git pull

if [ "$1" = "frontend" ]; then
  # Обновляем только фронтенд
  echo "Обновление только фронтенда..."
  cd frontend
  docker build -t ms_frontend .
  cd ..
  docker compose up -d --no-deps frontend
else
  # Обновляем всё
  echo "Обновление бэкенда и фронтенда..."
  cd backend
  docker build -t ms_backend .
  cd ..
  cd frontend
  docker build -t ms_frontend .
  cd ..
  docker compose down
  docker compose up -d
fi