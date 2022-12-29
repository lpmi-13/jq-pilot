import { Fragment, useState } from "react";
import "./App.css";

function isOpen(ws) {
    return ws.readyState === ws.OPEN;
}

function App() {
    const [wsQuestion, setWsQuestion] = useState(null);
    const [wsAnswer, setWsAnswer] = useState(null);

    const ws = new WebSocket(`${process.env.REACT_APP_ENV === "production" ? "wss" : "ws"}://localhost:8000/ws`);

    setInterval(() => {
        if (!isOpen(ws)) {
            return;
        }
        ws.send("update");
    }, 1000);

    ws.onmessage = ({ data }) => {
        const { question, answer } = JSON.parse(data);
        setWsQuestion(question);
        setWsAnswer(answer);
    };

    return (
        <Fragment>
            {wsQuestion ? (
                <Fragment>
                    <div className="App">
                        <h2>here's the question!</h2>
                    </div>
                    <div className="flexblock">
                        <div className="codeblock">
                            <pre>{JSON.stringify(wsQuestion, null, 2)}</pre>
                        </div>
                        <div className="arrow">{`=>`}</div>
                        <div className="codeblock">
                            <pre>{JSON.stringify(wsAnswer, null, 2)}</pre>
                        </div>
                    </div>
                    <div className="instructions">
                        Try to transform the structure from{" "}
                        <code>localhost:8000/question</code>
                        into the filtered data and send it to{" "}
                        <code>localhost:8000/answer</code>
                    </div>
                </Fragment>
            ) : (
                <h3 className="loading">LOADING...</h3>
            )}
        </Fragment>
    );
}

export default App;
