# nivogo
microservices for nivogo

docker build -t react-app .          
docker run -p 3000:3000 react-app

## Show tables
```bash
docker exec -it postgres-product psql -U postgres -d product
docker exec -it postgres-inventory psql -U postgres -d inventory
docker exec -it postgres-oms psql -U postgres -d oms
```
then
```psql
\dt
```
