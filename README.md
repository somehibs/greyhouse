# greyhouse

like grey matter but for houses.


*design*
apis take node keys
node keys are issued by a node management api
auxiliary data might be stored behind a node key

presence api for sending presense data
presence system for recieving presence data and determining room by room presence
allow for unknown presence and door entry/exit routines
allow for time of day/week to play a part in determining routines

also work on a light api for sending photosensitive resistor data to primary node
work on primary node response code that reacts to presence and light api calls
sensors probably shouldn't publish events more than every few seconds.

the presence api needs to handle guessing unknown visitors.