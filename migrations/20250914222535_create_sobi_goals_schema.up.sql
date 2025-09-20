CREATE TYPE goal_status AS ENUM ('active', 'completed');

CREATE TABLE missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    day_number INT NOT NULL, 
    focus VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL
);

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID NOT NULL,
    text TEXT NOT NULL,
    FOREIGN KEY (mission_id) REFERENCES missions(id) ON DELETE CASCADE
);

CREATE TABLE user_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    goal_category VARCHAR(100) NOT NULL,
    status goal_status DEFAULT 'active',
    start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    target_end_date DATE NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(uid) ON DELETE CASCADE
);

CREATE TABLE task_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_goal_id UUID NOT NULL,
    task_id UUID NOT NULL,
    user_id UUID NOT NULL, 
    is_completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP NULL,
    FOREIGN KEY (user_goal_id) REFERENCES user_goals(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(uid) ON DELETE CASCADE,
    UNIQUE (user_goal_id, task_id)
);

CREATE TABLE mission_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_goal_id UUID NOT NULL,
    mission_id UUID NOT NULL,
    user_id UUID NOT NULL, 
    is_completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMP NULL,
    total_tasks INT DEFAULT 0, 
    completed_tasks INT DEFAULT 0, 
    completion_percentage DECIMAL(5,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_goal_id) REFERENCES user_goals(id) ON DELETE CASCADE,
    FOREIGN KEY (mission_id) REFERENCES missions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(uid) ON DELETE CASCADE,
    UNIQUE (user_goal_id, mission_id)
);

CREATE TABLE goal_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_goal_id UUID NOT NULL,
    user_id UUID NOT NULL,
    goal_category VARCHAR(100) NOT NULL,
    total_days INT NOT NULL,
    days_completed INT NOT NULL,
    total_missions INT NOT NULL,
    missions_completed INT NOT NULL,
    total_tasks INT NOT NULL,
    tasks_completed INT NOT NULL,
    completion_percentage DECIMAL(5,2) NOT NULL,
    reflection VARCHAR(255),
    self_changes JSON,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_goal_id) REFERENCES user_goals(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(uid) ON DELETE CASCADE
);
