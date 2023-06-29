import React from "react";
import { AstroData } from "../types";

import SunriseImage from "../assets/sunrise.svg";
import SunsetImage from "../assets/sunset.svg";
import MoonriseImage from "../assets/moonrise.svg";
import MoonsetImage from "../assets/moonset.svg";

import {
    LocationInfo,
    EventCard,
    MoonPhaseCard,
    IlluminationCard,
} from "../components";

type AstroCardProps = {
    data: AstroData;
};

export const AstroCard: React.FC<AstroCardProps> = ({ data }) => (
    <div
        style={{
            display: "flex",
            flexWrap: "wrap",
            justifyContent: "space-between",
        }}
    >
        <LocationInfo
            name={data.name}
            region={data.region}
            country={data.country}
        />
        <p
            style={{ flexBasis: "100%", textAlign: "center" }}
        >{`Date: ${data.date}`}</p>
        <EventCard
            eventName="Sunrise"
            eventTime={data.astro.sunrise}
            image={SunriseImage}
        />
        <EventCard
            eventName="Sunset"
            eventTime={data.astro.sunset}
            image={SunsetImage}
        />
        <EventCard
            eventName="Moonrise"
            eventTime={data.astro.moonrise}
            image={MoonriseImage}
        />
        <EventCard
            eventName="Moonset"
            eventTime={data.astro.moonset}
            image={MoonsetImage}
        />
        <MoonPhaseCard phase={data.astro.moon_phase} />
        <IlluminationCard illumination={data.astro.moon_illumination} />
    </div>
);
