import React, { useState } from "react";
import { Layout, Menu } from "antd";
import { HomeOutlined, ReadOutlined, EditOutlined } from "@ant-design/icons";
import App from "../../App";

const { Sider, Footer, Content } = Layout;

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false);

  function onCollapse(input) {
    setCollapsed(input);
  }

  return (
    <>
      <Layout style={{ minHeight: "100vh" }}>
        <Sider collapsible collapsed={collapsed} onCollapse={onCollapse}>
          <div className="logo" />
          <Menu theme="dark" defaultSelectedKeys={["1"]} mode="inline">
            <Menu.Item key="1" icon={<HomeOutlined />}>
              Home
            </Menu.Item>
            <Menu.Item key="2" icon={<ReadOutlined />}>
              Read
            </Menu.Item>
            <Menu.Item key="3" icon={<EditOutlined />}>
              Write
            </Menu.Item>
          </Menu>
        </Sider>
        <Layout className="site-layout">
        <Content style={{ margin: '0 16px' }}>
            <App />
          </Content>
          <Footer style={{ textAlign: "center" }}>OkayFileSystem © 2022</Footer>
        </Layout>
      </Layout>
    </>
  );
}
