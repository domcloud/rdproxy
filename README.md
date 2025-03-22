# RDPROXY

This is a redis proxy to make ACL more convenient. It prefixes keys with the ACL username before sending it to upstream and undoing it when it about to send downstream.

This software is currently WIP. Only support RESP2.

Your app can connect to this instance listening by default at port `6479`. 
