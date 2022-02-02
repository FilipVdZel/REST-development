# REST-development
///////////////////////////////////////////////////////////////////////////////////
How to use webUsers:
Get All Users (GET):
    - Run #curl loscalhost:8081/users

Create new User (POST):
    - Run #curl -X POST  loscalhost:8081/users -d '{"name":"", "surname":"", "username":"", "password":"", "dob":""}'


Update User (PUT):
    - Run # curl localhost:8081/users/{id} -X PUT -d 'json with updated fields' |jq





///////////////////////////////////////////////////////////////////////////////////////////
How to use webSubscriptions:

Get Subscriptions (GET):
    - Run # curl localhost:8082/subscriptions |jq
    - Responce will be json documents of all subscription channels with name, owner, and discription

Create Subscription (POST):
    - Run # curl -X POST -- user Username:Password localhost:8082/subscriptions  -d '{"name":"name","description":"description"}
    - Will pass on username and password and validate it
    - API will respond with: "Channel 'name' was created by user 'User'

Update Subscription (PUT):
    - Run # curl -X PUT --user Username:Password localhost:8082/subscriptions/{id} -d    'Json with updated values'

Delete Subscription (DELETE):
    - Run # curl -X DELETE --user Username:Password localhost:8082/subscriptions/{id} 

Subscribe to Channel (POST):
    - Run # curl -X POST localhost:8082/subscribe/{id}?username
    - Will return text saying user subscribed successfully

Unsubscibe from Channel (DELETE):
    - Run # curl _X DELETE localhost:8082/subscribe/{id}?username
    - Will return text saying user unsubscribed successfully