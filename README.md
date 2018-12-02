# go-chat

A very simple chat server written in golang.

Supported commands:
```
* /who                              Prints current chat users
* /nick <newnick>                   Changes nick to <newnick> if it's not in use
* /nick <nick> <password>           Logs user in with specified nick and password
* /register <password>              Registers currently picked nick with given password
```
Current implementation:
* Generates a username of the form "GuestNNN" for newly connected users. Generated name collisions are currently not validated.
* Uses in-memory db for registered users.
* Sends all messages to all users, including the sender.
* Disconnects user if same user logs in using different client.

