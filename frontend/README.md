# JQ Pilot Frontend

This is the react frontend for the JQ Pilot project.

It's very minimal and basically just shows whatever data is coming from the backend. It starts up a Server Sent Events connection that only pushes data when the state chages (ie, somebody sends the correct answer).

The main reason for the push on change is so that if the backend is deployed somewhere public (currently at https://jkew.party), you the data shown to the user will update if somebody else successfully passes the expected answer to the server.

Because there's no interaction done in the frontend, there's no need for the frontend to send any messages to the server, and this is why I went with SSE (the original implementation was using websockets, which was wasteful, and this is a much cleaner implementation).
