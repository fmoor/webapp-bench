CREATE VIEW GraphQLUserDetails := (
    WITH MODULE default
    SELECT User {
        latest_reviews := (
            SELECT User.<author
            ORDER BY .creation_time DESC
        )
    }
);


CREATE VIEW GraphQLPersonDetails := (
    WITH MODULE default
    SELECT Person {
        acted_in := (
            SELECT Person.<cast
        ),
        directed := (
            SELECT Person.<directors
        ),
    }
);


CREATE VIEW GraphQLMovieDetails := (
    WITH MODULE default
    SELECT Movie {
        reviews := (
            SELECT Movie.<movie
        ),
    }
);