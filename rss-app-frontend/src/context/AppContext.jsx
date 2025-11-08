import React, { createContext, useState, useEffect, useContext } from 'react';

// Create the context
const AppContext = createContext();

// Create the Provider component
export const AppProvider = ({ children }) => {
    const [feeds, setFeeds] = useState([]);
    const [articles, setArticles] = useState([]);
    const [loading, setLoading] = useState(true); // To show loading status
    
    const [favoriteIds, setFavoriteIds] = useState(() => {
        const savedFavorites = localStorage.getItem('favoriteArticles');
        return savedFavorites ? JSON.parse(savedFavorites) : [];
    });
    
    useEffect(() => {
        const fetchInitialData = async () => {
            setLoading(true);
            try {
                // Use Promise.all to fetch feeds and articles in parallel
                const [feedsResponse, articlesResponse] = await Promise.all([
                    fetch('/api/feeds'),
                    fetch('/api/articles')
                ]);

                if (!feedsResponse.ok || !articlesResponse.ok) {
                    throw new Error('Network response was not ok');
                }

                const feedsData = await feedsResponse.json();
                const articlesData = await articlesResponse.json();
                
                setFeeds(feedsData);
                setArticles(articlesData);

            } catch (error) {
                console.error("Failed to fetch initial data:", error);
            } finally {
                setLoading(false);
            }
        };

        fetchInitialData();
    }, []);

    useEffect(() => {
        localStorage.setItem('favoriteArticles', JSON.stringify(favoriteIds));
    }, [favoriteIds]);

    // API calls

    const addFeed = async (url) => {
        if (!url) return;
        try {
            const response = await fetch('/api/feeds', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ url: url })
            });

            if (!response.ok) throw new Error('Failed to add feed');

            const feedsResponse = await fetch('/api/feeds');
            const feedsData = await feedsResponse.json();
            setFeeds(feedsData);

        } catch (error) {
            console.error("Error adding feed:", error);
        }
    };

    const removeFeed = async (url) => {
        try {
            const response = await fetch('/api/feeds', {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ url: url })
            });

            if (!response.ok) throw new Error('Failed to remove feed');

            // Update state by filtering out the removed feed
            setFeeds(feeds.filter(feedUrl => feedUrl !== url));

        } catch (error) {
            console.error("Error removing feed:", error);
        }
    };

    const toggleFavorite = (articleId) => {
        setFavoriteIds(prevIds => {
            if (prevIds.includes(articleId)) {
                return prevIds.filter(id => id !== articleId);
            } else {
                return [...prevIds, articleId];
            }
        });
    };

    const value = {
        feeds,
        addFeed,
        removeFeed,
        articles,
        setArticles,
        favoriteIds,
        toggleFavorite,
        loading
    };

    return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
};

export const useAppContext = () => {
    return useContext(AppContext);
};