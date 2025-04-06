# nivogo
Microservices and Fronted for checkout system for nivogo.
- OMS
- Inventory
- Product
- Frontend(React)

## How to Run
```sh
make up
```

## How does it work
- After using command above front end will fill cart with 2 items using api call to product-api.
- After pressing checkout 2 second loading screen will appear. Afterwards frontend will send request to oms-api to create an order.
- OMS built using SAGA pattern. If order fails it will revert back to the stock number of items.



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

## Check stocks before and after creating an order.
```sh
curl --location --request GET 'localhost:9081/inventory?ids=1%2C2' \
--header 'Content-Type: application/json' \
--data '{
    "name":"test3",
    "description":"test3",
    "price":1
}'
```

---
docker build -t react-app .          
docker run -p 3000:3000 react-app