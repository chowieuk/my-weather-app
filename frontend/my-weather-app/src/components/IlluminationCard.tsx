import React from "react";

type IlluminationCardProps = {
    illumination: number;
};

export const IlluminationCard: React.FC<IlluminationCardProps> = ({
    illumination,
}) => (
    <div>
        <h3>Moon Illumination</h3>
        <div style={{ backgroundColor: "grey", width: "100%", height: "20px" }}>
            <div
                style={{
                    backgroundColor: "yellow",
                    width: `${illumination}%`,
                    height: "20px",
                }}
            />
        </div>
        <p>{`${illumination}%`}</p>
    </div>
);
