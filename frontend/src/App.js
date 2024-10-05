import { Fragment, useEffect, useState } from 'react';
import { Formatter, FracturedJsonOptions } from 'fracturedjsonjs';
import './styles/App.scss';

const currentDomain =
    process.env.REACT_APP_ENV === 'production' &&
    !window.location.origin.endsWith('gitpod.io')
        ? window.location.origin
        : 'http://localhost:8000';

const formatter = new Formatter();
const options = new FracturedJsonOptions();
options.MaxTotalLineLength = 40;
options.IndentSpaces = 2;
formatter.Options = options;

function App() {
    const [question, setQuestion] = useState(null);
    const [answer, setAnswer] = useState(null);
    const [prompt, setPrompt] = useState(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const eventSource = new EventSource(`${currentDomain}/sse`);

        eventSource.onopen = (event) => {
            console.log('SSE connection opened:', event);
            setError(null);
        };

        eventSource.onmessage = (event) => {
            console.log({ event });
            console.log('Received SSE message:', event.data);
            try {
                const { answer, prompt, question } = JSON.parse(event.data);
                setAnswer(answer);
                setQuestion(question);
                setPrompt(prompt);
                setIsLoading(false);
            } catch (err) {
                console.error('Error parsing SSE message:', err);
                setError('Error parsing server message');
            }
        };

        eventSource.onerror = (error) => {
            console.error('EventSource failed:', error);
            setError(`SSE Error: ${error.message || 'Unknown error'}`);
            setIsLoading(true);
            eventSource.close();
        };

        return () => {
            console.log('Closing SSE connection');
            eventSource.close();
        };
    }, []);

    return (
        <Fragment>
            {error && <div className="error">{error}</div>}
            <div className={`center ${question ? 'visible' : 'invisible'}`}>
                <div>{prompt}</div>
            </div>
            <h3 className={`loading ${isLoading ? 'visible' : 'undisplay'}`}>
                LOADING...
            </h3>
            <div className={`flexblock ${question ? 'visible' : 'invisible'}`}>
                <div className="codeblock">
                    <pre>{formatter.Serialize(question)}</pre>
                </div>
                <div className="arrow">{`=>`}</div>
                <div className="codeblock">
                    <pre>
                        {typeof answer === 'string' ||
                        typeof answer === 'number'
                            ? answer
                            : formatter.Serialize(answer)}
                    </pre>
                </div>
            </div>
            <div
                className={`codeblock instructions ${
                    question ? 'visible' : 'invisible'
                }`}
            >
                Try to transform the structure from{' '}
                <pre>{currentDomain}/question</pre>
                into the filtered data and send it to{' '}
                <pre>{currentDomain}/answer</pre>
            </div>
        </Fragment>
    );
}

export default App;
