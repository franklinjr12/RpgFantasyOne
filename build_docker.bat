docker build -t singlefantasy-builder -f Dockerfile.builder .
docker create --name temp singlefantasy-builder
docker cp temp:/app/output/app.exe ./output/
docker rm temp