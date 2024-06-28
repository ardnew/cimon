# cimon
## Multi-layer service monitor/agent

This module provides a single command-line executable that [services](#services)
highly-concurrent inbound IP network connections. 
It supports two communication protocols, distinguished by the
[abstraction layer](https://en.wikipedia.org/wiki/OSI_model#Layer_architecture)
in which they operate (per [OSI reference model](https://en.wikipedia.org/wiki/OSI_model)):

|Protocol|Abstraction layer|Functional API|
|:------:|:---------------:|:-------------|
|  [TCP](https://en.wikipedia.org/wiki/Transmission_Control_Protocol)   | [Transport](https://en.wikipedia.org/wiki/OSI_model#Layer_4:_Transport_layer) | Raw byte stream, application-defined messages/encodings (bidirectional) |
|  [HTTP](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol)  | [Application](https://en.wikipedia.org/wiki/OSI_model#Layer_7:_Application_layer) | [REST](https://en.wikipedia.org/wiki/Web_API), [URI-based API](https://en.wikipedia.org/wiki/Web_API) (inbound, query) • [JSON](https://en.wikipedia.org/wiki/JSON), [YAML](https://en.wikipedia.org/wiki/YAML) (outbound, response) |

## Services

TBD
