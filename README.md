# nivogo
microservices for nivogo

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
