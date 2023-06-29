import React from "react";

type LocationInfoProps = {
    name: string;
    region: string;
    country: string;
};

export const LocationInfo: React.FC<LocationInfoProps> = ({
    name,
    region,
    country,
}) => (
    <p
        style={{ flexBasis: "100%", textAlign: "center" }}
    >{`${name}, ${region}, ${country}`}</p>
);
