import React from "react";
import { AstroData } from "../types";
import { LocationInfo, AstroDetailsCard } from "."; // Assuming you're exporting these components

type AstroCardProps = {
    data: AstroData;
};

export const AstroCard: React.FC<AstroCardProps> = ({ data }) => (
    <div>
        <LocationInfo
            name={data.name}
            region={data.region}
            country={data.country}
        />
        <p>{`Date: ${data.date}`}</p>
        <AstroDetailsCard details={data.astro} />
    </div>
);
