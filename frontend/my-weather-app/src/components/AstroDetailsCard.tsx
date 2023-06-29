import React from "react";
import { AstroDetails } from "../types";

type AstroDetailsProps = {
    details: AstroDetails;
};

export const AstroDetailsCard: React.FC<AstroDetailsProps> = ({ details }) => (
    <div>
        <p>{`Sunrise: ${details.sunrise}`}</p>
        <p>{`Sunset: ${details.sunset}`}</p>
        <p>{`Moonrise: ${details.moonrise}`}</p>
        <p>{`Moonset: ${details.moonset}`}</p>
        <p>{`Moon Phase: ${details.moon_phase}`}</p>
        <p>{`Moon Illumination: ${details.moon_illumination}`}</p>
    </div>
);
