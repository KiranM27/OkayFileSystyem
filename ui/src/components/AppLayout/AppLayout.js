import React, { useState } from "react";
import { Layout, Menu } from "antd";
import { HomeOutlined, ReadOutlined, EditOutlined, FolderOpenOutlined } from "@ant-design/icons";
import App from "../../App";
import "./AppLayout.css"

const { Sider, Header, Footer, Content } = Layout;

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false);

  function onCollapse(input) {
    setCollapsed(input);
  }

  return (
    <>
      <Layout style={{ minHeight: "100vh" }}>
        <Sider collapsible collapsed={collapsed} onCollapse={onCollapse}>
          <div className="logo" style= {{color: "white", textAlign: "center"}}> <FolderOpenOutlined /> </div>
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
        <Header style={{ padding: 0, color: "white", textAlign: "center" }} >  Okay File System </Header>
        <Content style={{ margin: '0 16px' }}>
            <App />
          </Content>
          <Footer style={{ textAlign: "center" }}>OkayFileSystem © 2022</Footer>
        </Layout>
      </Layout>
    </>
  );
}
