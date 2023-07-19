import { Fragment, useState } from "react";
import { Formatter, FracturedJsonOptions } from "fracturedjsonjs";
import "./styles/App.scss";

const isOpen = (ws) => {
    return ws.readyState === ws.OPEN;
};

const currentHost =
    process.env.REACT_APP_ENV === "production"
        ? window.location.host
        : "localhost:8000";

const ws = new WebSocket(
    `${
        process.env.REACT_APP_ENV === "production" ? "wss" : "ws"
    }://${currentHost}/ws`
);

const currentDomain =
    process.env.REACT_APP_ENV === "production"
        ? window.location.origin
        : "localhost:8000";

const formatter = new Formatter();
const options = new FracturedJsonOptions();
options.MaxTotalLineLength = 40;
options.IndentSpaces = 2;
formatter.Options = options;

function App() {
    const [wsQuestion, setWsQuestion] = useState(null);
    const [wsAnswer, setWsAnswer] = useState(null);
    const [wsPrompt, setWsPrompt] = useState(null);

    // probably just do this once on startup, and then after that just wait for a push
    setInterval(() => {
        if (!isOpen(ws)) {
            return;
        }
        ws.send("update");
    }, 2000);

    ws.onmessage = ({ data }) => {
        const { answer, prompt, question } = JSON.parse(data);
        setWsAnswer(answer);
        setWsQuestion(question);
        setWsPrompt(prompt);
    };

    return (
        <Fragment>
            <div className={`App ${wsQuestion ? "visible" : "invisible"}`}>
                <h2>{wsPrompt}</h2>
            </div>
            <h3 className={`loading ${wsQuestion ? "undisplay" : "visible"}`}>
                LOADING...
            </h3>
            <div
                className={`flexblock ${wsQuestion ? "visible" : "invisible"}`}
            >
                <div className="codeblock">
                    <pre>{formatter.Serialize(wsQuestion)}</pre>
                </div>
                <div className="arrow">{`=>`}</div>
                <div className="codeblock">
                    <pre>
                        {/* we need something smarter to determine if
                                we should ask for the user to pass a quoted string
                                or just the raw value, but this will do for now */}
                        {(typeof wsAnswer === "string") |
                        (typeof wsAnswer === "number")
                            ? wsAnswer
                            : formatter.Serialize(wsAnswer)}
                    </pre>
                </div>
            </div>
            <div
                className={`codeblock instructions ${
                    wsQuestion ? "visible" : "invisible"
                }`}
            >
                Try to transform the structure from{" "}
                <pre>{currentDomain}/question</pre>
                into the filtered data and send it to{" "}
                <pre>{currentDomain}/answer</pre>
            </div>
        </Fragment>
    );
}

export default App;
