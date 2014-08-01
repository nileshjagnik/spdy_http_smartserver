spdy_http_smartserver
=====================

A server that can detect, even without NPN, whether an incoming request is spdy or http and serve them accordingly.

Details
=======

This project is an extension to a high performance [Amahi spdy](https://github.com/amahi/spdy) library. Spdy is a high performance protocol that has been developed by google to overcome the shortcomings of http. Spdy clients conventionally use Next Protocol Negotiation (NPN) to convey to spdy-complaint servers that they want to use spdy. This project contains a server that can listen for both spdy and http without NPN. This is especially useful for developing back-end support using spdy for iOS applications since iOS does not have NPN in its current version. The server can obviously be used to serve connections that use NPN over SSL.
 
Install
=======
`go get github.com/nileshjagnik/spdy_http_smartserver`

Examples
========
See the examples folder for working examples
