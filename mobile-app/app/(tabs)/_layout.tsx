import { Tabs } from 'expo-router';
import { LayoutDashboard, Link2, User } from 'lucide-react-native';

export default function TabLayout() {
    return (
        <Tabs
            screenOptions={{
                headerShown: false,
                tabBarStyle: {
                    backgroundColor: '#18181b',
                    borderTopColor: '#27272a',
                    paddingTop: 8,
                    height: 85,
                },
                tabBarActiveTintColor: '#3b82f6',
                tabBarInactiveTintColor: '#a3a3a3',
                tabBarLabelStyle: {
                    fontSize: 12,
                    fontWeight: '500',
                    paddingBottom: 8,
                },
            }}
        >
            <Tabs.Screen
                name="index"
                options={{
                    title: 'Dashboard',
                    tabBarIcon: ({ color }) => <LayoutDashboard size={24} color={color} />,
                }}
            />
            <Tabs.Screen
                name="chains"
                options={{
                    title: 'Chains',
                    tabBarIcon: ({ color }) => <Link2 size={24} color={color} />,
                }}
            />
            <Tabs.Screen
                name="addresses"
                options={{
                    title: 'Addresses',
                    tabBarIcon: ({ color }) => <User size={24} color={color} />,
                }}
            />
        </Tabs>
    );
}
