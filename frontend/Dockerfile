# NOTE: the build context is the repository root (see docker-compose.yml),
# so the COPY paths below are relative to the repo root, not to frontend/.
FROM node:22-alpine

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm install

COPY frontend/ ./

ENTRYPOINT ["npm", "run", "dev"]
