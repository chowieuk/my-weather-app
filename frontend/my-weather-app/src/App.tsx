// src/App.tsx
import React, { useState } from "react";
import { AstroData } from "./types";
import { getAstroData } from "./utils";

function App() {
    const [location, setLocation] = useState("");
    const [data, setData] = useState<AstroData | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        const data = await getAstroData(location);
        setData(data);
    };

    return (
        <div className="App">
            <form onSubmit={handleSubmit}>
                <input
                    value={location}
                    onChange={(e) => setLocation(e.target.value)}
                    placeholder="Enter location"
                />
                <button type="submit">Submit</button>
            </form>
            {data && (
                <div>
                    <p>{`Date: ${data.date}`}</p>
                    <p>{`Sunrise: ${data.astro.sunrise}`}</p>
                    <p>{`Sunset: ${data.astro.sunset}`}</p>
                    <p>{`Moonrise: ${data.astro.moonrise}`}</p>
                    <p>{`Moonset: ${data.astro.moonset}`}</p>
                    <p>{`Moon Phase: ${data.astro.moon_phase}`}</p>
                    <p>{`Moon Illumination: ${data.astro.moon_illumination}`}</p>
                    <p>{`Expires at: ${data.expires_at}`}</p>
                </div>
            )}
        </div>
    );
}

export default App;
