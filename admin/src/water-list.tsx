import { Datagrid, DateField, List, NumberField, ReferenceField, TextField } from 'react-admin';

export const WaterList = () => (
    <List>
        <Datagrid rowClick="edit">
            <TextField source="id" />
            <TextField source="user_id"/>
            <NumberField source="amount" />
            <DateField source="created_at" />
            <DateField source="updated_at" />
        </Datagrid>
    </List>
);