## Basic concepts
This utility allows applying a group policy (eg. block internet) on WiFi devices. Think of a scenario when your child's Internet is shaped to dial-up speed until they have 90% results in the mathletics tests.

## Terminology
**Cell** - a loose collection of mobile devices that belong to the same household, eg. a household with 2 children, 3 laptops, 1 iPad, etc. These devices can be connected to the same WiFi network (eg. your house), or multiple (eg. grandparents' house with different network). You may want to try and enforce the policies consistently - in all the networks that are "within your reach". Eg. you don't want to try and enforce these policies on school routers etc.
**Cell Admin** - typically a parent who can control one or many cells. There can be multiple admins for a cell - eg. both parents and a grandpa.
**Router** - a single router that belongs to a cell (can belong to one cell only).
**Device** - a WiFi device that can connect to a router, typically this device belongs to a child.

We need to store network details in the database

## Bot interface
The utility has a bot interface for convenience. Built-in commands are as follows:
**/list** - list all the cells for the current user (cell admin)
**/newcell** - add a new cell
**/status** - list currently applied policy per device

Welcome message for new users: 
I haven't recognised you, you may need to register first    

Flow for new user: 
/newcell - create a new cell with name
