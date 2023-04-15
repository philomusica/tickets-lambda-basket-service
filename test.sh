#!/bin/bash

curl -X POST -H "Content-Type: application/json" https://api.philomusica.org.uk/order -d '
	{
		"emailAddress": "joshuacrunden@gmail.com",
		"firstName": "Joshua",
		"lastName": "Crunden",
		"orderLines": [
			{ 
				"concertId": "1044",
				"numOfFullPrice": 2,
				"numOfConcessions": 0
			}
		]
	}
'
