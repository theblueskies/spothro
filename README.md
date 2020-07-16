# spot

Prerequisites to start the service: Docker
To start the service:
`
make start
`

Run tests with:
`
make test
`
This will start the service on port 9000   


# What's included:  
1. Rates Service  
2. Dockerfile  
3. Additional Metrics endpoint  
4. Swagger Spec in swagger.yml  


# Description of rates service  
All initial rates are seeded from rates/seed_rates.json. When the rates are stored, they are first converted to their UTC time equivalents and then stored on the key of weekday.  

When a request comes in asking for a rate, the input time ranges are first converted to their UTC time equivalents and the rates are then looked up.   

The /metrics endpoint uses Prometheus to collect and output metrics on the GET and PUT of rates endpoints. It collects the count of type of responses, the average latency across all types of GET responses and PUT responses.  

There is no tight coupling between the router and API. This is enabled through the use of an interface.
The router expects an interface to be passed in. The API struct satisfies the Service interface. This enables the rates service not to be tied down to the sole implementation of rates service as defined in this problem statement(JSON inputs, or how it's stored). A new API can easily supersede and replace the existing API by simply implementing the Service interface.  

# Available endpoints:
1. GET /rate  
2. PUT /rates  
3. GET /health  
4. GET /metrics  

## Example requests:  
1. GET call needs to have the datetime parameters encoded
`
GET 127.0.0.1:9000/rate?start_time=2015-07-04T07%3A00%3A00%2B05%3A00&end_time=2015-07-04T20%3A00%3A00%2B05%3A00
`


2. PUT needs a body with the rates to update the rates on the service:  
Example:  

`
PUT /rates
{
    "rates": [
        {
            "days": "mon,tues,thurs",
            "times": "0900-2100",
            "tz": "America/Chicago",
            "price": 1500
        }
    ]
}
`
