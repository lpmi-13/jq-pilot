import { Fragment, useEffect, useState } from "react";
import "./App.css";

function App() {
    const ws = new WebSocket("ws://localhost:8000/");

    ws.onopen = () => ws.send("ping");

    // setInterval(() => ws.send("ping"), 1000);

    const [question, setQuestion] = useState([]);

    useEffect(() => {
        fetch("/question")
            .then((response) => response.json())
            .then((data) => setQuestion(data));
    }, []);

    return (
        <Fragment>
            <div className="App">
                <h2>here's the question!</h2>
            </div>
            <div>
                <h3>
                    {question.map(({ id, name, favoriteColors }) => (
                        <div key={id}>
                            <h3>Person</h3>
                            <div>ID: {id}</div>
                            <div>Name: {name}</div>
                            <div>Fave colors {favoriteColors.join()}</div>
                        </div>
                    ))}
                </h3>
            </div>
            <div>
                <h2>
                    Try to transform it into just the first element via jq and
                    send it back to the server at localhost:8000/answer
                </h2>
            </div>
        </Fragment>
    );
}

export default App;
