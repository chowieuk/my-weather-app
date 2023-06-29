// App.tsx
import React, { useState } from "react";
import { AstroData } from "./types";
import { getAstroData } from "./utils";
import { SearchForm, AstroCard } from "./components";

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
            <SearchForm
                location={location}
                setLocation={setLocation}
                onSubmit={handleSubmit}
            />
            {data && <AstroCard data={data} />}
        </div>
    );
}

export default App;
