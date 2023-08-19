import React, { useState, useEffect } from 'react'
import { createRoot } from 'react-dom/client'
import styled from 'styled-components'
import axios from 'axios'
import schema from '../../schema_pb'
import CpuProfileView from './cpu_profile_view'
import MemleakProfileView from './memleak_profile_view'

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
var routines = []

const TabContent = ({ routine }) => {
  switch (routine) {
    case 'cpu_profile':
      return <CpuProfileView />
    case 'memleak_profile':
      return <MemleakProfileView />
  }
  return null
}

const TabGroup = () => {
  const [active, setActive] = useState('')
  const [isLoading, setLoading] = useState(true);
  const fetchRoutines = async () => {
    let resp = await axios.get("/api/routines", { responseType: 'arraybuffer' })
    routines = schema.Routines.deserializeBinary(resp.data).getRoutinesList()
    setActive(routines[0])
    setLoading(false)
  }
  const tabTitle = (routine) => {
    switch (routine) {
      case 'cpu_profile':
        return "CPU Profile"
      case 'memleak_profile':
        return "Memleak Profile"
    }
    return ""
  }

  useEffect(() => {
    fetchRoutines()
  }, [])

  if (isLoading) {
    return <div>Loading...</div>
  }
  return (
    <div>
      <ButtonGroup>
        {routines.map(routine => (
          <Tab
            key={routine}
            active={active === routine}
            onClick={() => setActive(routine)}
          >
            {tabTitle(routine)}
          </Tab>
        ))}
      </ButtonGroup>
      <TabContent
        routine={active}
      />
    </div>
  )
}

const container = document.getElementById('root')
const root = createRoot(container)
root.render(<TabGroup />)
