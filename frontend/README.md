# JQ Pilot Frontend

This is the react frontend for the JQ Pilot project.

It's very minimal and basically just shows whatever data is coming from the backend. It starts up a websocket connect to the backend server and occassionally polls for data.

The main reason for the polling is so that if the backend is deployed somewhere public (currently at https://jkew.party), you the data shown to the user will update if somebody else successfully passes the expected answer to the server.

It's very possible this would also be achievable (and with fewer network requests) if the backend pushed websocket updates when things changed rather than the frontend needing to poll. Because there's no interaction done in the frontend, there's no need for the frontend to send any websocket messages to the server, so I'll probably update this to be more efficient down the line.
