-- +goose Up
CREATE TABLE payment_plans (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    image TEXT NOT NULL DEFAULT 'https://img.freepik.com/free-vector/subscriber-concept-illustration_114360-27827.jpg?t=st=1722672381~exp=1722675981~hmac=7f023cd2ad6bd97625fdeefdefa53d070ab5c4ed659db78d5570c883b5bddc27&w=740',
    description TEXT,
    duration TEXT NOT NULL default 'free',
    price DECIMAL(10, 2) NOT NULL,
    features TEXT[] NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'active'
);
-- Add our pre-made payment plans
INSERT INTO payment_plans (name, image, duration, price, features, description)
VALUES 
    ('Free', 
     'https://images.unsplash.com/photo-1547481887-a26e2cacb5b2?q=80&w=1470&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D', 
     'free', 
     0.00, 
     ARRAY[
         '5 Feeds: Track up to 5 of your favorite topics or websites.',
         'Follow 10 Feeds: Keep tabs on updates from 10 different feeds.',
         '10 Messages/Day: Send up to 10 personalized messages per day.',
         'Basic Analytics: Get insights into your feed consumption.',
         'Community Access: Join discussions and get tips from other users.'
     ], 
     'Get Started'),
    ('Monthly', 
     'https://images.unsplash.com/photo-1656941599882-808d7b04b86a?w=500&auto=format&fit=crop&q=60&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8Nnx8bW9udGhseXxlbnwwfHwwfHx8MA%3D%3D',
     'month', 
     10.00, 
     ARRAY[
         '20 Feeds: Aggregate up to 20 diverse feeds for a richer experience.',
         'Follow 40 Feeds: Stay informed with updates from 40 different feeds.',
         '50 Messages/Day: Engage more with up to 50 messages daily.',
         'Advanced Analytics: Detailed insights on your feed interactions and trends.',
         'Priority Support: Faster response times from our support team.',
         'Feed Scheduling: Schedule updates and content delivery at your convenience.',
         'Content Filtering: Filter content based on keywords or topics of interest.'
     ], 
     'Buy Now'),
    ('Annual',
     'https://images.unsplash.com/photo-1556742502-ec7c0e9f34b1?w=500&auto=format&fit=crop&q=60&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8MXx8eWVhcmx5JTIwcGF5bWVudHxlbnwwfHwwfHx8MA%3D%3D',
     'year', 
     100.00, 
     ARRAY[
         'Unlimited Feeds: Aggregate an unlimited number of feeds.',
         'Follow Unlimited Feeds: Never miss an update from any feed.',
         'Unlimited Messages/Day: Unlimited messaging capabilities for maximum engagement.',
         'Premium Analytics: Deep dive into trends, popular topics, and consumption patterns.',
         '24/7 Support: Get assistance any time of the day or night.',
         'Feed Customization: Personalize the appearance and organization of your feeds.',
         'Exclusive Content: Access to exclusive content and premium feed sources.',
         'Early Access to New Features: Be the first to try out new features and improvements.'
     ], 
     'Buy Now');



-- +goose Down
DROP TABLE payment_plans;