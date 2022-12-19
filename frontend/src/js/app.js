import React, { useState, useEffect } from 'react';
import { createRoot } from 'react-dom/client';
import styled from 'styled-components';
import axios from 'axios';
import schema from '../../schema_pb';

const Tab = styled.button`
  font-size: 20px;
  padding: 10px 60px;
  cursor: pointer;
  opacity: 0.6;
  background: white;
  border: 0;
  outline: 0;
  ${({ active }) =>
    active &&
    `
    border-bottom: 2px solid black;
    opacity: 1;
  `}
`;
const ButtonGroup = styled.div`
  display: flex;
`;
var tasks = []
const TabContent = (props) => {
  return <p> SELECT {props.task} </p>;
}

const TabGroup = () => {
  const [active, setActive] = useState('');

  useEffect(() => {
    const fetchTasks = async () => {
      let resp = await axios.get("/api/tasks", {responseType: 'arraybuffer'});
      tasks = schema.Tasks.deserializeBinary(resp.data).getTasksList()
      setActive(tasks[0])
    }

    fetchTasks()
  }, [])
  return (
    <>
      <ButtonGroup>
        {tasks.map(task => (
          <Tab
            key={task}
            active={active === task}
            onClick={() => setActive(task)}
          >
            {task}
          </Tab>
        ))}
      </ButtonGroup>
      <TabContent
        task={active}
      />
    </>
  );
}

const container = document.getElementById('root');
const root = createRoot(container);
root.render(<TabGroup />);