# cimon
## Concurrent TCP server with context

This module listens for incoming connections and will echo any newline-delimited
data received. It gracefully allows peer connections to be created and destroyed
concurrently, and it will close all connections after receiving certain 
interrupt signals or when commanded by a peer.

This serves as a template program capable of receiving and processing messages
from multiple hosts concurrently over a TCP/IP network. The immediate use case
is to coordinate iOS hardware provisioning for continuous integration and test
case automation.
