import React from "react";

type SearchFormProps = {
    location: string;
    setLocation: (location: string) => void;
    onSubmit: (event: React.FormEvent) => void;
};

export const SearchForm: React.FC<SearchFormProps> = ({
    location,
    setLocation,
    onSubmit,
}) => (
    <form
        style={{ flexBasis: "100%", textAlign: "center" }}
        onSubmit={onSubmit}
    >
        <input
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            placeholder="Enter location"
        />
        <button type="submit">Submit</button>
    </form>
);
