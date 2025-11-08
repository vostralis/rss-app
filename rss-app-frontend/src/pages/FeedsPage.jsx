import React, { useState } from 'react';
import { useAppContext } from '../context/AppContext';

const FeedsPage = () => {
    const { feeds, addFeed, removeFeed, setArticles } = useAppContext();
    const [updateStatus, setUpdateStatus] = useState('');

    const handleUpdateFeeds = async () => {
        setUpdateStatus('Updating feeds, please wait...');
        try {
            const response = await fetch('/api/articles/update', { method: 'POST' });
            if (!response.ok) {
                throw new Error('Update request failed');
            }
            const result = await response.json();
            setUpdateStatus(`Update complete! Found ${result.new_articles_count} new articles.`);

            const articlesResponse = await fetch('/api/articles');
            const articlesData = await articlesResponse.json();
            setArticles(articlesData);

        } catch (error) {
            console.error('Failed to update feeds:', error);
            setUpdateStatus('An error occurred during the update.');
        }
    };

    const handleAddFeed = () => {
        const newUrl = prompt('Enter the new RSS feed URL:');
        if (newUrl) {
            addFeed(newUrl.trim());
        }
    };

    return (
        <div className="page-content">
            <h1>My Feeds</h1>
            <div className="actions-container">
                <button onClick={handleAddFeed} className="action-btn">Add Feed</button>
                <button onClick={handleUpdateFeeds} className="action-btn">Fetch New Articles</button>
            </div>
            {updateStatus && <p className="status-message">{updateStatus}</p>}

            <div id="feeds-container">
                {feeds && feeds.length > 0 ? (
                    feeds.map(url => (
                        <div key={url} className="feed-item">
                            <span>{url}</span>
                            <button onClick={() => removeFeed(url)} className="remove-feed-btn">Remove</button>
                        </div>
                    ))
                ) : (
                    <p>No feeds added yet.</p>
                )}
            </div>
        </div>
    );
};

export default FeedsPage;