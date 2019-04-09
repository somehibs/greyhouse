# greyhouse

like grey matter but for houses.


*design*
ensure that all apis accept node keys which can be resolved down to the arbitrary node and it's configuration/info/etc

start with a node api for accepting node configuration tied to a node identifier, returns a node key
next work on a presence api for sending presence information to primary node
also work on a light api for sending photosensitive resistor data to primary node
work on primary node response code that reacts to presence and light api calls
sensors probably shouldn't publish events more than every few seconds.

the presence api needs to handle guessing unknown visitors.
